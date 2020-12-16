package com.webank.wecube.platform.core.repository.workflow;

import java.util.List;

import org.apache.ibatis.annotations.Param;
import org.springframework.stereotype.Repository;

import com.webank.wecube.platform.core.entity.workflow.ProcInstInfoEntity;
import com.webank.wecube.platform.core.entity.workflow.ProcInstInfoQueryEntity;

@Repository
public interface ProcInstInfoMapper{
    int deleteByPrimaryKey(Integer id);
    int insert(ProcInstInfoEntity record);
    int insertSelective(ProcInstInfoEntity record);
    ProcInstInfoEntity selectByPrimaryKey(Integer id);
    int updateByPrimaryKeySelective(ProcInstInfoEntity record);
    int updateByPrimaryKey(ProcInstInfoEntity record);
    
    /**
     * 
     * @param procInstKernelId
     * @return
     */
    ProcInstInfoEntity selectOneByProcInstKernelId(@Param("procInstKernelId") String procInstKernelId);

    /**
     * 
     * @param procDefIds
     * @return
     */
    List<ProcInstInfoEntity> selectByProcDefIdIn(@Param("procDefIds") List<String> procDefIds);
    
    /**
     * 
     * @param procDefId
     * @return
     */
    List<ProcInstInfoEntity> selectAllByProcDefId(@Param("procDefId") String procDefId);
    
    /**
     * 
     * @param roleNames
     * @return
     */
    List<ProcInstInfoQueryEntity> selectAllByProcInstInfoByRoleNames(@Param("roleNames")List<String> roleNames);
    
}
